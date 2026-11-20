/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x180159910 */

longlong FUN_180159910(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  longlong *plVar2;
  longlong lVar3;
  int *piVar4;
  longlong *plVar5;
  longlong *plVar6;
  longlong *plVar7;
  longlong *plVar8;
  longlong lVar9;
  code *pcVar10;
  code *pcVar11;
  int iVar12;
  longlong lVar13;
  longlong lVar14;
  undefined8 extraout_XMM0_Qa;
  undefined8 extraout_XMM0_Qa_00;
  undefined8 extraout_XMM0_Qa_01;
  undefined8 extraout_XMM0_Qa_02;
  undefined8 uVar15;
  undefined8 extraout_XMM0_Qa_03;
  longlong local_res18;
  undefined8 local_res20;
  undefined8 local_50;
  longlong *local_48;

  plVar5 = DAT_1803b6ba8;
  plVar1 = (longlong *)*param_3;
  plVar2 = (longlong *)param_3[1];
  local_res18 = DAT_180391d20;
  if (DAT_1803b6ba8 == (longlong *)0x0) {
LAB_18015996a:
    DAT_1803b6ba8 = (longlong *)FUN_18036cd10(param_1,DAT_1803b6b60,DAT_1803ca758,0x10);
  }
  else if ((1 < *DAT_1803b6ba8) || (DAT_1803b6ba8[2] != 0)) {
    if ((DAT_1803b6ba8 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b6ba8 && (*DAT_1803b6ba8 = *DAT_1803b6ba8 + -1, *plVar5 == 0)))) {
      (**(code **)(plVar5[1] + 0x30))(plVar5);
    }
    goto LAB_18015996a;
  }
  plVar5 = DAT_1803b6ba8;
  FUN_180001d00(param_1,DAT_1803b6ba8);
  plVar6 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803b6f60);
  if ((plVar6 == (longlong *)0x0) || (lVar9 = *plVar6, lVar9 == 0)) {
    PyErr_PrintEx(0);
    Py_Exit(1);
    pcVar11 = (code *)swi(3);
    lVar9 = (*pcVar11)();
    return lVar9;
  }
  lVar13 = 0;
  lVar3 = *(longlong *)(DAT_1803ca760 + 0x20);
  local_res20 = DAT_1803b6f68;
  if (*(char *)(lVar3 + 10) == '\0') {
    plVar6 = (longlong *)FUN_1800015a0(DAT_1803ca760,DAT_1803b6f70);
    uVar15 = extraout_XMM0_Qa_02;
    if (plVar6 != (longlong *)0x0) {
      lVar13 = *plVar6;
    }
LAB_180159abc:
    if (lVar13 == 0) goto LAB_180159ac1;
LAB_180159b05:
    plVar6 = (longlong *)FUN_18037f0c0(uVar15,lVar13,DAT_1803b6f78);
    if (plVar6 == (longlong *)0x0) {
LAB_180159c83:
      lVar9 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
    }
    else {
      plVar7 = (longlong *)FUN_180381a40(plVar1,plVar2);
      if (plVar7 != (longlong *)0x0) {
        if ((code *)plVar7[1] == PyFloat_Type_exref) {
          plVar8 = (longlong *)FUN_180371690(*(double *)(DAT_1803b6f80 + 0x10) * (double)plVar7[2]);
        }
        else {
          plVar8 = (longlong *)FUN_1803817e0(plVar7);
        }
        if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
          (**(code **)(plVar7[1] + 0x30))(plVar7);
        }
        if (plVar8 != (longlong *)0x0) {
          *(undefined4 *)(plVar5 + 5) = 0x121;
          plVar7 = (longlong *)FUN_180376770(param_1,plVar6,plVar8);
          if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
            (**(code **)(plVar6[1] + 0x30))(plVar6);
          }
          if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
            (**(code **)(plVar8[1] + 0x30))(plVar8);
          }
          if (plVar7 != (longlong *)0x0) {
            plVar6 = (longlong *)FUN_180380b80(plVar7,DAT_1803b6f88);
            if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
              (**(code **)(plVar7[1] + 0x30))(plVar7);
            }
            if (plVar6 != (longlong *)0x0) {
              local_50 = local_res20;
              *(undefined4 *)(plVar5 + 5) = 0x121;
              local_48 = plVar6;
              lVar9 = FUN_180377750(param_1,lVar9,&local_50);
              if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                (**(code **)(plVar6[1] + 0x30))(plVar6);
              }
              if (lVar9 != 0) {
                FUN_180001d80(param_1);
                if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
                  (**(code **)(plVar1[1] + 0x30))(plVar1);
                }
                if ((int)*plVar2 < 0) {
                  return lVar9;
                }
                *plVar2 = *plVar2 + -1;
                if (*plVar2 != 0) {
                  return lVar9;
                }
                (**(code **)(plVar2[1] + 0x30))(plVar2);
                return lVar9;
              }
            }
          }
          goto LAB_180159c83;
        }
      }
      lVar9 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
        (**(code **)(plVar6[1] + 0x30))(plVar6);
      }
    }
  }
  else {
    iVar12 = *(int *)(lVar3 + 0xc);
    if (iVar12 == 0) {
      iVar12 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
      *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar12 + 1;
      *(int *)(lVar3 + 0xc) = iVar12;
    }
    uVar15 = extraout_XMM0_Qa;
    if (iVar12 != DAT_1803a9910) {
      DAT_1803a9910 = iVar12;
      DAT_1803b6be0 = FUN_1803707d0(lVar3,DAT_1803b6f70,*(undefined8 *)(DAT_1803b6f70 + 0x18));
      uVar15 = extraout_XMM0_Qa_00;
    }
    if (-1 < DAT_1803b6be0) {
      lVar14 = (1L << (*(byte *)(lVar3 + 9) & 0x3f)) + lVar3;
      lVar13 = *(longlong *)(lVar14 + 0x28 + DAT_1803b6be0 * 0x10);
      if (lVar13 != 0) goto LAB_180159b05;
      DAT_1803b6be0 = FUN_1803707d0(lVar3,DAT_1803b6f70,*(undefined8 *)(DAT_1803b6f70 + 0x18));
      if (-1 < DAT_1803b6be0) {
        lVar13 = *(longlong *)(lVar14 + 0x28 + DAT_1803b6be0 * 0x10);
        uVar15 = extraout_XMM0_Qa_01;
        goto LAB_180159abc;
      }
    }
LAB_180159ac1:
    plVar6 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803b6f70);
    if ((plVar6 != (longlong *)0x0) && (lVar13 = *plVar6, uVar15 = extraout_XMM0_Qa_03, lVar13 != 0)
       ) goto LAB_180159b05;
    FUN_180373340(param_1,&local_res18,DAT_1803b6f70);
    lVar9 = local_res18;
  }
  pcVar11 = (code *)0x0;
  piVar4 = *(int **)(lVar9 + 0x28);
  if (piVar4 == (int *)0x0) {
    pcVar10 = (code *)FUN_18038e2a0(plVar5,0x121);
    plVar6 = *(longlong **)(lVar9 + 0x28);
    if (pcVar10 != _Py_NoneStruct_exref) {
LAB_180159cf0:
      pcVar11 = pcVar10;
      pcVar10 = pcVar11;
      if ((pcVar11 != (code *)0x0) && (*(int *)pcVar11 + 1 != 0)) {
        *(int *)pcVar11 = *(int *)pcVar11 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar4 + 6) == plVar5) goto LAB_180159d39;
    pcVar10 = (code *)FUN_18038e2a0(plVar5,0x121);
    *(int **)(pcVar10 + 0x10) = piVar4;
    if (*piVar4 + 1 != 0) {
      *piVar4 = *piVar4 + 1;
    }
    plVar6 = *(longlong **)(lVar9 + 0x28);
    if (pcVar10 != _Py_NoneStruct_exref) goto LAB_180159cf0;
  }
  *(code **)(lVar9 + 0x28) = pcVar11;
  if (((plVar6 != (longlong *)0x0) && (-1 < (int)*plVar6)) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)
     ) {
    (**(code **)(plVar6[1] + 0x30))();
  }
  if (((pcVar10 != (code *)0x0) && (-1 < *(int *)pcVar10)) &&
     (*(longlong *)pcVar10 = *(longlong *)pcVar10 + -1, *(longlong *)pcVar10 == 0)) {
    (**(code **)(*(longlong *)(pcVar10 + 8) + 0x30))(pcVar10);
  }
LAB_180159d39:
  FUN_18036d0a0(plVar5,&DAT_180391be4,plVar1,plVar2);
  plVar6 = DAT_1803b6ba8;
  if (plVar5 == DAT_1803b6ba8) {
    if ((-1 < (int)*DAT_1803b6ba8) && (*DAT_1803b6ba8 = *DAT_1803b6ba8 + -1, *plVar6 == 0)) {
      (**(code **)(plVar6[1] + 0x30))();
    }
    DAT_1803b6ba8 = (longlong *)0x0;
  }
  lVar3 = **(longlong **)(param_1 + 0x38);
  plVar5 = *(longlong **)(lVar3 + 0x30);
  **(longlong **)(param_1 + 0x38) = *(longlong *)(lVar3 + 8);
  *(undefined4 *)(plVar5 + 8) = 0xffffffff;
  plVar6 = (longlong *)plVar5[2];
  if (((plVar6 != (longlong *)0x0) && (plVar5[2] = 0, -1 < (int)*plVar6)) &&
     (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
    (**(code **)(plVar6[1] + 0x30))();
  }
  if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
    (**(code **)(plVar5[1] + 0x30))(plVar5);
  }
  plVar5[10] = 0;
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
    (**(code **)(plVar2[1] + 0x30))(plVar2);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar9;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return 0;
}
