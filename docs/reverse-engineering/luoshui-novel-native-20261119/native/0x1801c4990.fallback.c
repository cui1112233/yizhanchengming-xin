/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1801c4990 */

longlong FUN_1801c4990(longlong param_1,longlong param_2,longlong *param_3)

{
  undefined8 *puVar1;
  int *piVar2;
  code *pcVar3;
  undefined8 uVar4;
  char cVar5;
  int iVar6;
  longlong *plVar7;
  longlong lVar8;
  longlong *plVar9;
  undefined4 uVar10;
  int local_res10 [2];
  longlong *local_res20;
  longlong *local_68;
  longlong *local_60;
  longlong *local_58;
  longlong *local_50;
  undefined8 local_48;
  undefined8 local_40;
  undefined8 local_38 [2];
  undefined4 uVar11;
  
  plVar7 = DAT_1803bbde8;
  uVar4 = DAT_180391dc8;
  puVar1 = *(undefined8 **)(param_2 + 0x90);
  if (*(int *)(param_2 + 0x78) == 1) {
    FUN_180371770((longlong)puVar1 + 0x24,&local_58,8,local_38,8,local_res10,4,&local_res20,8,
                  &local_50,8,&local_60,8,&local_40,8,&local_68,8,&local_48,8,0);
    if (param_3 == (longlong *)0x0) {
      uVar11 = 0x70;
      goto LAB_1801c4df2;
    }
    if ((-1 < (int)*param_3) && (*param_3 = *param_3 + -1, *param_3 == 0)) {
      (**(code **)(param_3[1] + 0x30))(param_3);
    }
    cVar5 = FUN_18000d5b0(param_1);
    if (cVar5 != '\0') goto LAB_1801c4a7d;
LAB_1801c4ba2:
    uVar11 = 0x6f;
LAB_1801c4df2:
    uVar10 = uVar11;
    puVar1[3] = *(undefined8 *)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
  }
  else {
    *puVar1 = 0;
    puVar1[1] = 0;
    puVar1[2] = 0;
    puVar1[3] = uVar4;
    *(undefined4 *)(puVar1 + 4) = 0;
    if (plVar7 == (longlong *)0x0) {
LAB_1801c4a17:
      plVar7 = (longlong *)FUN_18036cd10(param_1,DAT_1803bbd80,DAT_1803ca610,0x10);
      DAT_1803bbde8 = plVar7;
    }
    else if ((1 < *plVar7) || (plVar7[2] != 0)) {
      if ((plVar7 != (longlong *)0x0) &&
         ((-1 < (int)*plVar7 && (*plVar7 = *plVar7 + -1, *plVar7 == 0)))) {
        (**(code **)(plVar7[1] + 0x30))();
      }
      goto LAB_1801c4a17;
    }
    *(longlong **)(param_2 + 0x50) = plVar7;
    iVar6 = (int)*plVar7 + 1;
    if (iVar6 != 0) {
      *(int *)plVar7 = iVar6;
    }
    lVar8 = *(longlong *)(param_2 + 0x50);
    *(longlong *)(lVar8 + 0x38) = param_2;
    *(undefined4 *)(lVar8 + 0x40) = 0;
    FUN_180001df0(param_1,*(undefined8 *)(param_2 + 0x50));
    piVar2 = *(int **)(param_1 + 0xf8);
    *(int **)(param_2 + 0x68) = piVar2;
    if ((piVar2 != (int *)0x0) && (*piVar2 + 1 != 0)) {
      *piVar2 = *piVar2 + 1;
    }
LAB_1801c4a7d:
    lVar8 = *(longlong *)(*(longlong *)(*(longlong *)(param_2 + 0xa0) + 0x10) + 8);
    pcVar3 = *(code **)(lVar8 + 0xe0);
    if (pcVar3 == (code *)0x0) {
      PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,"\'%s\' object is not an iterator",
                   *(undefined8 *)(lVar8 + 0x18));
LAB_1801c4bb7:
      cVar5 = FUN_18000cd20(param_1,*(undefined8 *)PyExc_StopIteration_exref);
      if (cVar5 != '\0') {
        plVar7 = *(longlong **)(param_2 + 0x68);
        if (((plVar7 != (longlong *)0x0) && (*(undefined8 *)(param_2 + 0x68) = 0, -1 < (int)*plVar7)
            ) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
          (**(code **)(plVar7[1] + 0x30))();
        }
        plVar7 = (longlong *)puVar1[1];
        if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) &&
           (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
          (**(code **)(plVar7[1] + 0x30))();
        }
        plVar7 = (longlong *)*puVar1;
        puVar1[1] = 0;
        if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) &&
           (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
          (**(code **)(plVar7[1] + 0x30))();
        }
        *puVar1 = 0;
        return 0;
      }
      goto LAB_1801c4ba2;
    }
    lVar8 = (*pcVar3)();
    if (lVar8 == 0) goto LAB_1801c4bb7;
    plVar7 = (longlong *)puVar1[1];
    puVar1[1] = lVar8;
    if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) &&
       (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))();
    }
    piVar2 = (int *)puVar1[1];
    plVar7 = (longlong *)*puVar1;
    *puVar1 = piVar2;
    if (*piVar2 + 1 != 0) {
      *piVar2 = *piVar2 + 1;
    }
    if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) &&
       (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))();
    }
    local_40 = *puVar1;
    local_60 = (longlong *)FUN_18037f0c0();
    uVar11 = 0x70;
    uVar10 = 0x70;
    if (local_60 == (longlong *)0x0) goto LAB_1801c4df2;
    *(undefined4 *)(*(longlong *)(param_2 + 0x50) + 0x28) = 0x70;
    local_res20 = (longlong *)FUN_1803770d0(param_1,local_60,DAT_1803bbec0);
    if ((-1 < (int)*local_60) && (*local_60 = *local_60 + -1, *local_60 == 0)) {
      (**(code **)(local_60[1] + 0x30))();
    }
    if (local_res20 == (longlong *)0x0) goto LAB_1801c4df2;
    local_res10[0] = FUN_180001510(local_res20);
    if (local_res10[0] != -1) {
      plVar7 = local_res20;
      if (local_res10[0] == 1) {
LAB_1801c4db5:
        local_38[0] = DAT_1803bbf20;
        local_58 = plVar7;
        lVar8 = FUN_180383520(plVar7);
        if ((-1 < (int)*local_58) && (*local_58 = *local_58 + -1, *local_58 == 0)) {
          (**(code **)(local_58[1] + 0x30))();
        }
        if (lVar8 != 0) {
          FUN_180371700((longlong)puVar1 + 0x24,&local_58,8,local_38,8,local_res10,4,&local_res20,8,
                        &local_50,8,&local_60,8,&local_40,8,&local_68,8,&local_48,8,0);
          *(undefined4 *)(param_2 + 0x78) = 1;
          return lVar8;
        }
      }
      else {
        if ((-1 < (int)*local_res20) && (*local_res20 = *local_res20 + -1, *local_res20 == 0)) {
          (**(code **)(local_res20[1] + 0x30))();
        }
        local_48 = *puVar1;
        local_68 = (longlong *)FUN_18037f0c0();
        if (local_68 != (longlong *)0x0) {
          *(undefined4 *)(*(longlong *)(param_2 + 0x50) + 0x28) = 0x70;
          local_50 = (longlong *)FUN_1803770d0(param_1,local_68,DAT_1803bbec8);
          if ((-1 < (int)*local_68) && (*local_68 = *local_68 + -1, *local_68 == 0)) {
            (**(code **)(local_68[1] + 0x30))();
          }
          plVar7 = local_50;
          if (local_50 != (longlong *)0x0) goto LAB_1801c4db5;
        }
      }
      goto LAB_1801c4df2;
    }
    puVar1[3] = *(undefined8 *)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
    if ((-1 < (int)*local_res20) && (*local_res20 = *local_res20 + -1, *local_res20 == 0)) {
      (**(code **)(local_res20[1] + 0x30))();
    }
  }
  plVar9 = puVar1 + 3;
  *(undefined4 *)(puVar1 + 4) = uVar10;
  puVar1[2] = &DAT_180391c08;
  plVar7 = (longlong *)puVar1[1];
  *(undefined4 *)(puVar1 + 0x86) = uVar10;
  *(undefined4 *)(puVar1 + 4) = 0;
  puVar1[0x85] = *plVar9;
  *plVar9 = 0;
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))();
  }
  *plVar9 = puVar1[0x85];
  *(undefined4 *)(puVar1 + 4) = *(undefined4 *)(puVar1 + 0x86);
  puVar1[1] = 0;
  cVar5 = FUN_180001380(param_1,plVar9);
  if (cVar5 != '\0') goto LAB_1801c4eef;
  piVar2 = *(int **)(*plVar9 + 0x28);
  if (piVar2 == (int *)0x0) {
    lVar8 = FUN_18038e2a0(*(undefined8 *)(param_2 + 0x50));
LAB_1801c4ea8:
    FUN_180001300(plVar9,lVar8);
  }
  else if ((*(int *)(puVar1 + 4) != 0) &&
          (*(longlong *)(piVar2 + 6) != *(longlong *)(param_2 + 0x50))) {
    lVar8 = FUN_18038e2a0();
    *(int **)(lVar8 + 0x10) = piVar2;
    if (*piVar2 + 1 != 0) {
      *piVar2 = *piVar2 + 1;
    }
    goto LAB_1801c4ea8;
  }
  FUN_18036d0a0(*(undefined8 *)(param_2 + 0x50),puVar1[2],0,*puVar1);
  plVar7 = DAT_1803bbde8;
  if (*(longlong **)(param_2 + 0x50) == DAT_1803bbde8) {
    if ((-1 < (int)*DAT_1803bbde8) && (*DAT_1803bbde8 = *DAT_1803bbde8 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))();
    }
    DAT_1803bbde8 = (longlong *)0x0;
  }
LAB_1801c4eef:
  plVar7 = *(longlong **)(param_2 + 0x68);
  if (((plVar7 != (longlong *)0x0) && (*(undefined8 *)(param_2 + 0x68) = 0, -1 < (int)*plVar7)) &&
     (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
    (**(code **)(plVar7[1] + 0x30))();
  }
  plVar7 = (longlong *)*puVar1;
  *(undefined4 *)(puVar1 + 0x88) = *(undefined4 *)(puVar1 + 4);
  puVar1[0x87] = *plVar9;
  *(undefined4 *)(puVar1 + 4) = 0;
  *plVar9 = 0;
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))();
  }
  *plVar9 = puVar1[0x87];
  *puVar1 = 0;
  *(undefined4 *)(puVar1 + 4) = *(undefined4 *)(puVar1 + 0x88);
  plVar7 = *(longlong **)(param_1 + 0x60);
  *(undefined8 *)(param_1 + 0x60) = puVar1[0x87];
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))();
  }
  return 0;
}

