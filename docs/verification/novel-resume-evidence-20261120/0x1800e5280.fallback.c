/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1800e5280 */

longlong FUN_1800e5280(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  longlong *plVar2;
  longlong *plVar3;
  longlong lVar4;
  longlong *plVar5;
  undefined8 uVar6;
  longlong *plVar7;
  longlong *plVar8;
  code *pcVar9;
  code *pcVar10;
  int iVar11;
  longlong *plVar12;
  longlong lVar13;
  undefined4 uVar14;
  longlong *plVar15;
  longlong *local_res18;
  longlong *local_res20;
  undefined8 local_70;
  longlong *local_68;
  longlong *local_60;
  longlong *local_58;
  undefined8 local_50;
  longlong *local_48;

  plVar5 = DAT_1803b3b20;
  plVar1 = (longlong *)*param_3;
  plVar2 = (longlong *)param_3[1];
  plVar12 = (longlong *)0x0;
  plVar15 = (longlong *)0x0;
  plVar3 = (longlong *)param_3[2];
  local_res20 = (longlong *)0x0;
  if (DAT_1803b3b20 == (longlong *)0x0) {
LAB_1800e52e6:
    DAT_1803b3b20 = (longlong *)FUN_18036cd10(param_1,DAT_1803b3ce0,DAT_1803ca778,0x28);
  }
  else if ((1 < *DAT_1803b3b20) || (DAT_1803b3b20[2] != 0)) {
    if ((DAT_1803b3b20 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b3b20 && (*DAT_1803b3b20 = *DAT_1803b3b20 + -1, *plVar5 == 0)))) {
      (**(code **)(plVar5[1] + 0x30))(plVar5);
    }
    goto LAB_1800e52e6;
  }
  plVar5 = DAT_1803b3b20;
  FUN_180001d00(param_1,DAT_1803b3b20);
  *(undefined4 *)(plVar5 + 5) = 0x79;
  plVar7 = (longlong *)FUN_18037e560(param_1,plVar1,DAT_1803b3f68);
  uVar6 = DAT_1803b3fd0;
  if (plVar7 == (longlong *)0x0) {
    lVar13 = *(longlong *)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
    plVar15 = plVar12;
    uVar14 = 0x79;
  }
  else {
    *(undefined4 *)(plVar5 + 5) = 0x7a;
    local_res18 = (longlong *)CONCAT44(local_res18._4_4_,0x7a);
    local_70 = uVar6;
    local_res20 = plVar7;
    local_68 = plVar2;
    plVar8 = (longlong *)FUN_18037e820(param_1,plVar7,DAT_1803b3f70,&local_70);
    if (plVar8 == (longlong *)0x0) {
      lVar13 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      plVar15 = plVar12;
      uVar14 = local_res18._0_4_;
    }
    else {
      iVar11 = FUN_180001510(plVar8);
      if (iVar11 == -1) {
        lVar13 = *(longlong *)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
          (**(code **)(plVar8[1] + 0x30))(plVar8);
        }
        plVar15 = plVar12;
        uVar14 = local_res18._0_4_;
      }
      else {
        if (iVar11 != 1) {
          if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
            (**(code **)(plVar8[1] + 0x30))(plVar8);
          }
          uVar14 = local_res18._0_4_;
          *(undefined4 *)(plVar5 + 5) = local_res18._0_4_;
          local_60 = plVar3;
          local_58 = plVar2;
          plVar8 = (longlong *)FUN_18037e820(param_1,plVar7,DAT_1803b3f70,&local_60);
          if (plVar8 == (longlong *)0x0) {
            lVar13 = *(longlong *)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
            goto LAB_1800e54df;
          }
        }
        plVar15 = (longlong *)FUN_18037f0c0();
        uVar14 = 0x7b;
        if (plVar15 != (longlong *)0x0) {
          local_50 = DAT_1803b3fe0;
          *(undefined4 *)(plVar5 + 5) = 0x7b;
          local_res18 = plVar3;
          local_48 = plVar8;
          local_res18 = (longlong *)
                        FUN_18037d650(param_1,plVar15,&local_50,&local_res18,DAT_1803b3fe8);
          if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
            (**(code **)(plVar15[1] + 0x30))(plVar15);
          }
          if (local_res18 != (longlong *)0x0) {
            FUN_180001d80(param_1);
            if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
              (**(code **)(plVar7[1] + 0x30))(plVar7);
            }
            if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
              (**(code **)(plVar8[1] + 0x30))(plVar8);
            }
            if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
              (**(code **)(plVar1[1] + 0x30))(plVar1);
            }
            if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
              (**(code **)(plVar2[1] + 0x30))(plVar2);
            }
            if ((int)*plVar3 < 0) {
              return (longlong)local_res18;
            }
            *plVar3 = *plVar3 + -1;
            if (*plVar3 != 0) {
              return (longlong)local_res18;
            }
            (**(code **)(plVar3[1] + 0x30))(plVar3);
            return (longlong)local_res18;
          }
        }
        lVar13 = *(longlong *)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        plVar15 = plVar8;
      }
    }
  }
LAB_1800e54df:
  local_res18 = *(longlong **)(lVar13 + 0x28);
  if (local_res18 == (longlong *)0x0) {
    pcVar9 = (code *)FUN_18038e2a0(plVar5,uVar14);
    plVar12 = *(longlong **)(lVar13 + 0x28);
    if (pcVar9 == _Py_NoneStruct_exref) {
      pcVar10 = (code *)0x0;
    }
    else {
      pcVar10 = pcVar9;
      if ((pcVar9 != (code *)0x0) && (*(int *)pcVar9 + 1 != 0)) {
        *(int *)pcVar9 = *(int *)pcVar9 + 1;
      }
    }
  }
  else {
    if ((longlong *)local_res18[3] == plVar5) goto LAB_1800e55a3;
    pcVar9 = (code *)FUN_18038e2a0(plVar5,uVar14);
    *(longlong **)(pcVar9 + 0x10) = local_res18;
    iVar11 = (int)*local_res18 + 1;
    if (iVar11 != 0) {
      *(int *)local_res18 = iVar11;
    }
    plVar12 = *(longlong **)(lVar13 + 0x28);
    if (pcVar9 == _Py_NoneStruct_exref) {
      pcVar10 = (code *)0x0;
    }
    else {
      pcVar10 = pcVar9;
      if ((pcVar9 != (code *)0x0) && (*(int *)pcVar9 + 1 != 0)) {
        *(int *)pcVar9 = *(int *)pcVar9 + 1;
      }
    }
  }
  *(code **)(lVar13 + 0x28) = pcVar10;
  if (((plVar12 != (longlong *)0x0) && (-1 < (int)*plVar12)) &&
     (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
    (**(code **)(plVar12[1] + 0x30))();
  }
  if (((pcVar9 != (code *)0x0) && (-1 < *(int *)pcVar9)) &&
     (*(longlong *)pcVar9 = *(longlong *)pcVar9 + -1, *(longlong *)pcVar9 == 0)) {
    (**(code **)(*(longlong *)(pcVar9 + 8) + 0x30))(pcVar9);
  }
LAB_1800e55a3:
  plVar7 = local_res20;
  FUN_18036d0a0(plVar5,"ooooo",plVar1,plVar2,plVar3,local_res20,plVar15);
  plVar12 = DAT_1803b3b20;
  if (plVar5 == DAT_1803b3b20) {
    if ((-1 < (int)*DAT_1803b3b20) && (*DAT_1803b3b20 = *DAT_1803b3b20 + -1, *plVar12 == 0)) {
      (**(code **)(plVar12[1] + 0x30))();
    }
    DAT_1803b3b20 = (longlong *)0x0;
  }
  lVar4 = **(longlong **)(param_1 + 0x38);
  plVar5 = *(longlong **)(lVar4 + 0x30);
  **(longlong **)(param_1 + 0x38) = *(longlong *)(lVar4 + 8);
  *(undefined4 *)(plVar5 + 8) = 0xffffffff;
  plVar12 = (longlong *)plVar5[2];
  if (((plVar12 != (longlong *)0x0) && (plVar5[2] = 0, -1 < (int)*plVar12)) &&
     (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
    (**(code **)(plVar12[1] + 0x30))();
  }
  if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
    (**(code **)(plVar5[1] + 0x30))(plVar5);
  }
  plVar5[10] = 0;
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))(plVar7);
  }
  if (((plVar15 != (longlong *)0x0) && (-1 < (int)*plVar15)) &&
     (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
    (**(code **)(plVar15[1] + 0x30))(plVar15);
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
    (**(code **)(plVar2[1] + 0x30))(plVar2);
  }
  if ((-1 < (int)*plVar3) && (*plVar3 = *plVar3 + -1, *plVar3 == 0)) {
    (**(code **)(plVar3[1] + 0x30))(plVar3);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar13;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return 0;
}
