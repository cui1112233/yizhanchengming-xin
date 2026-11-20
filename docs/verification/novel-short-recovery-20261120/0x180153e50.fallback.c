/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x180153e50 */

/* WARNING: Function: __security_check_cookie replaced with injection: security_check_cookie */

longlong FUN_180153e50(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  undefined8 *puVar1;
  int *piVar2;
  char cVar3;
  int iVar4;
  longlong lVar5;
  longlong lVar6;
  longlong *plVar7;
  longlong *plVar8;
  longlong *plVar9;
  longlong *plVar10;
  longlong *plVar11;
  undefined8 uVar12;
  longlong *plVar13;
  code *pcVar14;
  code *pcVar15;
  longlong lVar16;
  longlong lVar17;
  longlong *plVar18;
  undefined4 uVar19;
  undefined1 auStack_1c8 [32];
  longlong *local_1a8;
  longlong *local_1a0;
  longlong *local_198;
  longlong *local_190;
  longlong *local_188;
  longlong *local_180;
  longlong *local_178;
  longlong *local_170;
  longlong *local_168;
  longlong *local_160;
  longlong *local_158;
  longlong *local_150;
  longlong *local_140;
  longlong *local_138;
  longlong *local_130;
  longlong *local_128;
  longlong *local_120;
  undefined4 local_118;
  longlong *local_110;
  longlong *local_108;
  longlong *local_100;
  longlong *local_f8;
  undefined8 local_f0;
  longlong *local_e8;
  undefined8 local_e0;
  longlong *local_d8;
  longlong *local_d0;
  longlong *local_c8;
  longlong *local_c0;
  longlong *local_b8;
  longlong *local_b0;
  longlong *local_a8;
  longlong *local_a0;
  longlong *local_98;
  longlong *local_90;
  longlong *local_88;
  longlong *local_80;
  longlong *local_78;
  longlong *local_70;
  longlong *local_68;
  longlong *local_60;
  longlong *local_58;
  longlong *local_50;
  ulonglong local_48;

  plVar11 = DAT_1803b6b78;
  local_48 = DAT_1803a6040 ^ (ulonglong)auStack_1c8;
  local_108 = (longlong *)param_3[1];
  plVar10 = (longlong *)0x0;
  plVar9 = (longlong *)*param_3;
  plVar18 = (longlong *)0x0;
  local_168 = DAT_180391d20;
  local_138 = (longlong *)0x0;
  local_130 = (longlong *)0x0;
  local_128 = (longlong *)0x0;
  local_178 = (longlong *)0x0;
  local_120 = (longlong *)0x0;
  local_110 = (longlong *)0x0;
  local_100 = (longlong *)0x0;
  local_140 = DAT_1803b6b78;
  local_170 = plVar9;
  if (DAT_1803b6b78 == (longlong *)0x0) {
LAB_180153ef9:
    DAT_1803b6b78 = (longlong *)FUN_18036cd10(param_1,DAT_1803b6a40,DAT_1803ca758,0x40);
  }
  else if ((1 < *DAT_1803b6b78) || (DAT_1803b6b78[2] != 0)) {
    if ((DAT_1803b6b78 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b6b78 && (*DAT_1803b6b78 = *DAT_1803b6b78 + -1, *plVar11 == 0)))) {
      (**(code **)(plVar11[1] + 0x30))(plVar11);
    }
    goto LAB_180153ef9;
  }
  plVar11 = DAT_1803b6b78;
  local_140 = DAT_1803b6b78;
  FUN_180001d00(param_1,DAT_1803b6b78);
  lVar5 = FUN_18014c240(param_1);
  if (lVar5 == 0) {
    FUN_180373340(param_1,&local_168,DAT_1803b6d68);
    uVar19 = 0xab;
    plVar13 = local_168;
  }
  else {
    lVar6 = FUN_18014c020(param_1);
    if (lVar6 == 0) {
      FUN_180373340(param_1,&local_168,DAT_1803b6da0);
      uVar19 = 0xac;
      plVar13 = local_168;
      plVar10 = plVar18;
    }
    else {
      uVar19 = 0xac;
      local_80 = DAT_1803b6db0;
      *(undefined4 *)(plVar11 + 5) = 0xac;
      local_88 = plVar9;
      plVar7 = (longlong *)FUN_180377750(param_1,lVar6,&local_88);
      if (plVar7 == (longlong *)0x0) {
        plVar13 = *(longlong **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        plVar10 = plVar18;
      }
      else {
        iVar4 = FUN_180001510(plVar7);
        if (iVar4 == -1) {
LAB_180153ff1:
          plVar13 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          plVar10 = plVar18;
          if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
            (**(code **)(plVar7[1] + 0x30))(plVar7);
          }
        }
        else if (iVar4 == 1) {
LAB_1801540da:
          uVar19 = 0xab;
          *(undefined4 *)(plVar11 + 5) = 0xab;
          plVar8 = (longlong *)FUN_180376770(param_1,lVar5,plVar7);
          if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
            (**(code **)(plVar7[1] + 0x30))(plVar7);
          }
          if (plVar8 == (longlong *)0x0) {
LAB_180155b4a:
            plVar13 = *(longlong **)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
            plVar9 = local_170;
            plVar10 = plVar18;
          }
          else {
            iVar4 = FUN_180001510(plVar8);
            if (iVar4 == -1) {
              plVar13 = *(longlong **)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
              plVar9 = local_170;
              if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, plVar10 = plVar18, *plVar8 == 0))
              {
                (**(code **)(plVar8[1] + 0x30))(plVar8);
                plVar9 = local_170;
              }
            }
            else {
              if (iVar4 != 1) {
                if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                  (**(code **)(plVar8[1] + 0x30))(plVar8);
                }
                local_130 = DAT_1803b6dc0;
                iVar4 = (int)*DAT_1803b6dc0 + 1;
                plVar8 = local_130;
                if (iVar4 != 0) {
                  *(int *)DAT_1803b6dc0 = iVar4;
                }
              }
              local_130 = plVar8;
              lVar5 = FUN_18014c020(param_1);
              plVar9 = local_170;
              uVar12 = DAT_1803b6dd8;
              if (lVar5 != 0) {
                uVar19 = 0xb1;
                *(undefined4 *)(plVar11 + 5) = 0xb1;
                local_f8 = local_170;
                local_f0 = uVar12;
                plVar7 = (longlong *)FUN_180377750(param_1,lVar5,&local_f8);
                if (plVar7 == (longlong *)0x0) goto LAB_180155b4a;
                iVar4 = FUN_180001510(plVar7);
                if (iVar4 == -1) {
                  plVar13 = *(longlong **)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  plVar9 = local_170;
                  plVar10 = plVar18;
                  if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                    (**(code **)(plVar7[1] + 0x30))(plVar7);
                    plVar9 = local_170;
                  }
                  goto LAB_180155b57;
                }
                if (iVar4 != 1) {
                  if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                    (**(code **)(plVar7[1] + 0x30))(plVar7);
                  }
                  lVar5 = FUN_18014c020(param_1);
                  if (lVar5 == 0) {
                    FUN_180373340(param_1,&local_168,DAT_1803b6da0);
                    uVar19 = 0xb2;
                    plVar13 = local_168;
                    plVar9 = local_170;
                    plVar10 = plVar18;
                    goto LAB_180155b57;
                  }
                  uVar19 = 0xb2;
                  local_e0 = DAT_1803b6de0;
                  *(undefined4 *)(plVar11 + 5) = 0xb2;
                  local_e8 = plVar9;
                  plVar7 = (longlong *)FUN_180377750(param_1,lVar5,&local_e8);
                  if (plVar7 == (longlong *)0x0) goto LAB_180155b4a;
                }
                local_128 = plVar7;
                local_138 = (longlong *)FUN_180372160(param_1,0);
                uVar12 = DAT_1803ab258;
                plVar9 = (longlong *)FUN_180380b80(local_108,DAT_1803ab258);
                if (plVar9 != (longlong *)0x0) {
                  plVar7 = (longlong *)FUN_180375000(param_1,uVar12,plVar9,DAT_1803ab258);
                  if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                    (**(code **)(plVar9[1] + 0x30))(plVar9);
                  }
                  if (plVar7 != (longlong *)0x0) {
                    plVar9 = (longlong *)FUN_18000d280();
                    local_158 = plVar9;
                    if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                      (**(code **)(plVar7[1] + 0x30))(plVar7);
                    }
                    if (plVar9 != (longlong *)0x0) {
                      local_160 = (longlong *)(**(code **)(plVar9[1] + 0xe0))(plVar9);
                      if (local_160 != (longlong *)0x0) {
                        local_118 = 0xb9;
                        do {
                          plVar11 = local_120;
                          plVar9 = local_178;
                          plVar10 = local_160;
                          if (((local_100 != (longlong *)0x0) &&
                              (plVar10 = local_160, -1 < (int)*local_100)) &&
                             (*local_100 = *local_100 + -1, plVar10 = local_160, *local_100 == 0)) {
                            plVar10 = local_100 + 1;
                            local_100 = local_160;
                            (**(code **)(*plVar10 + 0x30))();
                            plVar10 = local_100;
                          }
                          local_100 = plVar10;
                          iVar4 = (int)*local_160 + 1;
                          if (iVar4 != 0) {
                            *(int *)local_160 = iVar4;
                          }
                          local_178 = local_160;
                          if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) &&
                             (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                            (**(code **)(plVar9[1] + 0x30))();
                          }
                          lVar5 = *(longlong *)(DAT_1803ca760 + 0x20);
                          if (*(char *)(lVar5 + 10) == '\0') {
                            plVar9 = (longlong *)FUN_1800015a0(DAT_1803ca760,DAT_1803b6de8);
                            if (plVar9 == (longlong *)0x0) {
                              lVar6 = 0;
                            }
                            else {
                              lVar6 = *plVar9;
                            }
LAB_18015449f:
                            if (lVar6 == 0) goto LAB_1801544a4;
                          }
                          else {
                            iVar4 = *(int *)(lVar5 + 0xc);
                            if (iVar4 == 0) {
                              iVar4 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
                              *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar4 + 1;
                              *(int *)(lVar5 + 0xc) = iVar4;
                            }
                            if (iVar4 != DAT_1803a992c) {
                              DAT_1803a992c = iVar4;
                              DAT_1803b6bd8 =
                                   FUN_1803707d0(lVar5,DAT_1803b6de8,
                                                 *(undefined8 *)(DAT_1803b6de8 + 0x18));
                            }
                            if (-1 < DAT_1803b6bd8) {
                              lVar16 = (1L << (*(byte *)(lVar5 + 9) & 0x3f)) + lVar5;
                              lVar6 = *(longlong *)(lVar16 + 0x28 + DAT_1803b6bd8 * 0x10);
                              if (lVar6 != 0) goto LAB_1801544cd;
                              DAT_1803b6bd8 =
                                   FUN_1803707d0(lVar5,DAT_1803b6de8,
                                                 *(undefined8 *)(DAT_1803b6de8 + 0x18));
                              if (-1 < DAT_1803b6bd8) {
                                lVar6 = *(longlong *)(lVar16 + 0x28 + DAT_1803b6bd8 * 0x10);
                                goto LAB_18015449f;
                              }
                            }
LAB_1801544a4:
                            plVar9 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803b6de8);
                            if (plVar9 == (longlong *)0x0) {
                              lVar6 = 0;
                            }
                            else {
                              lVar6 = *plVar9;
                            }
                            if (lVar6 == 0) {
                              FUN_180373340(param_1,&local_168,DAT_1803b6de8);
                              uVar19 = 0xb7;
                              plVar13 = local_168;
                              plVar10 = local_178;
                              goto LAB_1801553c2;
                            }
                          }
LAB_1801544cd:
                          lVar5 = FUN_18014c020(param_1);
                          piVar2 = DAT_1803b6d00;
                          if (lVar5 == 0) {
                            FUN_180373340(param_1,&local_168,DAT_1803b6da0);
                            uVar19 = 0xb8;
                            plVar13 = local_168;
                            plVar10 = local_178;
                            goto LAB_1801553c2;
                          }
                          if (local_170 == (longlong *)0x0) {
                            plVar9 = (longlong *)
                                     PyUnicode_FromFormat
                                               ("cannot access local variable \'%U\' where it is not associated with a value"
                                                ,DAT_1803b6df0);
                            plVar13 = (longlong *)
                                      FUN_180376770(param_1,*(undefined8 *)
                                                             PyExc_UnboundLocalError_exref,plVar9);
                            if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                              (**(code **)(plVar9[1] + 0x30))(plVar9);
                            }
                            FUN_18000d1d0(param_1,plVar13);
                            plVar10 = local_178;
                            uVar19 = 0xb8;
                            goto LAB_1801553c2;
                          }
                          lVar16 = *(longlong *)(param_1 + 0x10);
                          plVar9 = *(longlong **)(lVar16 + 0x41da8);
                          if (plVar9 == (longlong *)0x0) {
                            plVar9 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,3);
                          }
                          else {
                            *(longlong *)(lVar16 + 0x41da8) = plVar9[3];
                            *(int *)(lVar16 + 0x41e40) = *(int *)(lVar16 + 0x41e40) + -1;
                            *plVar9 = 1;
                          }
                          plVar10 = plVar9 + -2;
                          plVar9[3] = 0;
                          plVar9[4] = 0;
                          plVar9[5] = 0;
                          lVar16 = _PyThreadState_GetCurrent();
                          plVar18 = local_160;
                          lVar16 = *(longlong *)(*(longlong *)(lVar16 + 0x10) + 0xd0);
                          puVar1 = *(undefined8 **)(lVar16 + 8);
                          *puVar1 = plVar10;
                          plVar9[-1] = plVar9[-1] & 3;
                          plVar9[-1] = plVar9[-1] | (ulonglong)puVar1;
                          *plVar10 = lVar16;
                          *(longlong **)(lVar16 + 8) = plVar10;
                          if (*piVar2 + 1 != 0) {
                            *piVar2 = *piVar2 + 1;
                          }
                          plVar9[3] = (longlong)piVar2;
                          lVar16 = FUN_180375520(param_1,local_160,DAT_1803b6c50);
                          if (lVar16 == 0) {
                            uVar19 = 0xb8;
LAB_18015564b:
                            plVar13 = *(longlong **)(param_1 + 0x60);
                            *(undefined8 *)(param_1 + 0x60) = 0;
                            plVar10 = local_178;
                            if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                              (**(code **)(plVar9[1] + 0x30))(plVar9);
                              plVar10 = local_178;
                            }
                            goto LAB_1801553c2;
                          }
                          plVar9[4] = lVar16;
                          piVar2 = DAT_1803b6d08;
                          if (*DAT_1803b6d08 + 1 != 0) {
                            *DAT_1803b6d08 = *DAT_1803b6d08 + 1;
                          }
                          plVar9[5] = (longlong)piVar2;
                          plVar10 = (longlong *)PyUnicode_Join(DAT_1803b6c50,plVar9);
                          if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                            (**(code **)(plVar9[1] + 0x30))(plVar9);
                          }
                          if (plVar10 == (longlong *)0x0) {
LAB_180155757:
                            plVar13 = *(longlong **)(param_1 + 0x60);
                            *(undefined8 *)(param_1 + 0x60) = 0;
                            plVar10 = local_178;
                            uVar19 = 0xb8;
                            goto LAB_1801553c2;
                          }
                          *(undefined4 *)(local_140 + 5) = 0xb8;
                          local_d8 = local_170;
                          local_d0 = plVar10;
                          plVar9 = (longlong *)FUN_180377750(param_1,lVar5,&local_d8);
                          if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                            (**(code **)(plVar10[1] + 0x30))(plVar10);
                          }
                          if (plVar9 == (longlong *)0x0) goto LAB_180155757;
                          iVar4 = FUN_180001510(plVar9);
                          if (iVar4 == -1) {
                            plVar13 = *(longlong **)(param_1 + 0x60);
                            *(undefined8 *)(param_1 + 0x60) = 0;
                            if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                              (**(code **)(plVar9[1] + 0x30))(plVar9);
                            }
                            plVar10 = local_178;
                            uVar19 = 0xb8;
                            goto LAB_1801553c2;
                          }
                          if (iVar4 != 1) {
                            if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                              (**(code **)(plVar9[1] + 0x30))(plVar9);
                            }
                            lVar5 = *(longlong *)(DAT_1803ca760 + 0x20);
                            if (*(char *)(lVar5 + 10) == '\0') {
                              plVar9 = (longlong *)FUN_1800015a0();
                              if (plVar9 == (longlong *)0x0) {
                                lVar16 = 0;
                              }
                              else {
                                lVar16 = *plVar9;
                              }
LAB_18015472a:
                              if (lVar16 == 0) goto LAB_18015472f;
                            }
                            else {
                              iVar4 = *(int *)(lVar5 + 0xc);
                              if (iVar4 == 0) {
                                iVar4 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
                                *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar4 + 1;
                                *(int *)(lVar5 + 0xc) = iVar4;
                              }
                              if (iVar4 != DAT_1803a98dc) {
                                DAT_1803a98dc = iVar4;
                                DAT_1803b6bc0 =
                                     FUN_1803707d0(lVar5,DAT_1803b6df8,
                                                   *(undefined8 *)(DAT_1803b6df8 + 0x18));
                              }
                              if (-1 < DAT_1803b6bc0) {
                                lVar17 = (1L << (*(byte *)(lVar5 + 9) & 0x3f)) + lVar5;
                                lVar16 = *(longlong *)(lVar17 + 0x28 + DAT_1803b6bc0 * 0x10);
                                if (lVar16 != 0) goto LAB_180154757;
                                DAT_1803b6bc0 =
                                     FUN_1803707d0(lVar5,DAT_1803b6df8,
                                                   *(undefined8 *)(DAT_1803b6df8 + 0x18));
                                if (-1 < DAT_1803b6bc0) {
                                  lVar16 = *(longlong *)(lVar17 + 0x28 + DAT_1803b6bc0 * 0x10);
                                  goto LAB_18015472a;
                                }
                              }
LAB_18015472f:
                              plVar9 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803b6df8);
                              if ((plVar9 == (longlong *)0x0) || (lVar16 = *plVar9, lVar16 == 0)) {
                                FUN_180373340(param_1,&local_168,DAT_1803b6df8);
                                plVar13 = local_168;
                                plVar10 = local_178;
                                uVar19 = local_118;
                                goto LAB_1801553c2;
                              }
                            }
LAB_180154757:
                            uVar19 = local_118;
                            local_c0 = plVar18;
                            *(undefined4 *)(local_140 + 5) = local_118;
                            local_c8 = local_170;
                            plVar9 = (longlong *)FUN_180377750(param_1,lVar16,&local_c8);
                            if (plVar9 == (longlong *)0x0) {
                              plVar13 = *(longlong **)(param_1 + 0x60);
                              *(undefined8 *)(param_1 + 0x60) = 0;
                              plVar10 = local_178;
                              goto LAB_1801553c2;
                            }
                            iVar4 = FUN_180001510(plVar9);
                            if (iVar4 == -1) goto LAB_18015564b;
                            if (iVar4 != 1) {
                              if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                (**(code **)(plVar9[1] + 0x30))(plVar9);
                              }
                              piVar2 = DAT_1803b6d70;
                              lVar5 = *(longlong *)(param_1 + 0x10);
                              plVar18 = *(longlong **)(lVar5 + 0x41da8);
                              if (plVar18 == (longlong *)0x0) {
                                plVar18 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,3);
                              }
                              else {
                                *(longlong *)(lVar5 + 0x41da8) = plVar18[3];
                                *(int *)(lVar5 + 0x41e40) = *(int *)(lVar5 + 0x41e40) + -1;
                                *plVar18 = 1;
                              }
                              plVar9 = plVar18 + -2;
                              plVar18[3] = 0;
                              plVar18[4] = 0;
                              plVar18[5] = 0;
                              lVar5 = _PyThreadState_GetCurrent();
                              lVar5 = *(longlong *)(*(longlong *)(lVar5 + 0x10) + 0xd0);
                              puVar1 = *(undefined8 **)(lVar5 + 8);
                              *puVar1 = plVar9;
                              plVar18[-1] = plVar18[-1] & 3;
                              plVar18[-1] = plVar18[-1] | (ulonglong)puVar1;
                              *plVar9 = lVar5;
                              *(longlong **)(lVar5 + 8) = plVar9;
                              if (*piVar2 + 1 != 0) {
                                *piVar2 = *piVar2 + 1;
                              }
                              plVar18[3] = (longlong)piVar2;
                              lVar5 = FUN_180375520(param_1,local_160,DAT_1803b6c50);
                              if (lVar5 == 0) {
                                plVar13 = *(longlong **)(param_1 + 0x60);
                                *(undefined8 *)(param_1 + 0x60) = 0;
                                uVar19 = 0xba;
                                plVar10 = local_178;
                                if ((-1 < (int)*plVar18) &&
                                   (*plVar18 = *plVar18 + -1, *plVar18 == 0)) {
                                  (**(code **)(plVar18[1] + 0x30))(plVar18);
                                  plVar10 = local_178;
                                }
                                goto LAB_1801553c2;
                              }
                              plVar18[4] = lVar5;
                              piVar2 = DAT_1803b6d78;
                              if (*DAT_1803b6d78 + 1 != 0) {
                                *DAT_1803b6d78 = *DAT_1803b6d78 + 1;
                              }
                              plVar18[5] = (longlong)piVar2;
                              plVar9 = (longlong *)PyUnicode_Join(DAT_1803b6c50,plVar18);
                              if ((-1 < (int)*plVar18) && (*plVar18 = *plVar18 + -1, *plVar18 == 0))
                              {
                                (**(code **)(plVar18[1] + 0x30))(plVar18);
                              }
                              plVar18 = local_160;
                              if (plVar9 == (longlong *)0x0) {
                                plVar13 = *(longlong **)(param_1 + 0x60);
                                *(undefined8 *)(param_1 + 0x60) = 0;
                                plVar10 = local_178;
                                uVar19 = 0xba;
                                goto LAB_1801553c2;
                              }
                            }
                          }
                          *(undefined4 *)(local_140 + 5) = 0xb7;
                          local_b8 = plVar9;
                          local_b0 = plVar18;
                          local_150 = (longlong *)FUN_180377750(param_1,lVar6,&local_b8);
                          if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                            (**(code **)(plVar9[1] + 0x30))(plVar9);
                          }
                          if (local_150 == (longlong *)0x0) {
                            plVar13 = *(longlong **)(param_1 + 0x60);
                            *(undefined8 *)(param_1 + 0x60) = 0;
                            plVar10 = local_178;
                            uVar19 = 0xb7;
                            goto LAB_1801553c2;
                          }
                          local_120 = local_150;
                          if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
                             (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                            (**(code **)(plVar11[1] + 0x30))(plVar11);
                          }
                          lVar6 = DAT_1803ca760;
                          lVar5 = *(longlong *)(DAT_1803ca760 + 0x20);
                          if (*(char *)(lVar5 + 10) == '\0') {
                            plVar9 = (longlong *)FUN_1800015a0(DAT_1803ca760,DAT_1803b6e00);
                            lVar6 = DAT_1803ca760;
                            if (plVar9 == (longlong *)0x0) {
                              lVar16 = 0;
                            }
                            else {
                              lVar16 = *plVar9;
                            }
LAB_1801549fd:
                            if (lVar16 == 0) goto LAB_180154a02;
                          }
                          else {
                            iVar4 = *(int *)(lVar5 + 0xc);
                            if (iVar4 == 0) {
                              iVar4 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
                              *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar4 + 1;
                              *(int *)(lVar5 + 0xc) = iVar4;
                            }
                            if (iVar4 != DAT_1803a9908) {
                              DAT_1803a9908 = iVar4;
                              DAT_1803b6b20 =
                                   FUN_1803707d0(lVar5,DAT_1803b6e00,
                                                 *(undefined8 *)(DAT_1803b6e00 + 0x18));
                            }
                            if (-1 < DAT_1803b6b20) {
                              lVar17 = (1L << (*(byte *)(lVar5 + 9) & 0x3f)) + lVar5;
                              lVar16 = *(longlong *)(lVar17 + 0x28 + DAT_1803b6b20 * 0x10);
                              if (lVar16 != 0) goto LAB_180154a32;
                              DAT_1803b6b20 =
                                   FUN_1803707d0(lVar5,DAT_1803b6e00,
                                                 *(undefined8 *)(DAT_1803b6e00 + 0x18));
                              if (-1 < DAT_1803b6b20) {
                                lVar16 = *(longlong *)(lVar17 + 0x28 + DAT_1803b6b20 * 0x10);
                                goto LAB_1801549fd;
                              }
                            }
LAB_180154a02:
                            plVar9 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803b6e00);
                            if (plVar9 == (longlong *)0x0) {
                              lVar16 = 0;
                            }
                            else {
                              lVar16 = *plVar9;
                            }
                            lVar6 = DAT_1803ca760;
                            if (lVar16 == 0) {
                              FUN_180373340(param_1,&local_168,DAT_1803b6e00);
                              plVar13 = local_168;
                              plVar10 = local_178;
                              uVar19 = 0xbd;
                              goto LAB_1801553c2;
                            }
                          }
LAB_180154a32:
                          lVar5 = *(longlong *)(lVar6 + 0x20);
                          if (*(char *)(lVar5 + 10) == '\0') {
                            plVar9 = (longlong *)FUN_1800015a0(lVar6,DAT_1803b6e08);
                            if (plVar9 == (longlong *)0x0) {
                              lVar6 = 0;
                            }
                            else {
                              lVar6 = *plVar9;
                            }
LAB_180154af6:
                            if (lVar6 == 0) goto LAB_180154afb;
                          }
                          else {
                            iVar4 = *(int *)(lVar5 + 0xc);
                            if (iVar4 == 0) {
                              iVar4 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
                              *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar4 + 1;
                              *(int *)(lVar5 + 0xc) = iVar4;
                            }
                            if (iVar4 != DAT_1803a9920) {
                              DAT_1803a9920 = iVar4;
                              DAT_1803b6b00 =
                                   FUN_1803707d0(lVar5,DAT_1803b6e08,
                                                 *(undefined8 *)(DAT_1803b6e08 + 0x18));
                            }
                            if (-1 < DAT_1803b6b00) {
                              lVar17 = (1L << (*(byte *)(lVar5 + 9) & 0x3f)) + lVar5;
                              lVar6 = *(longlong *)(lVar17 + 0x28 + DAT_1803b6b00 * 0x10);
                              if (lVar6 != 0) goto LAB_180154b23;
                              DAT_1803b6b00 =
                                   FUN_1803707d0(lVar5,DAT_1803b6e08,
                                                 *(undefined8 *)(DAT_1803b6e08 + 0x18));
                              if (-1 < DAT_1803b6b00) {
                                lVar6 = *(longlong *)(lVar17 + 0x28 + DAT_1803b6b00 * 0x10);
                                goto LAB_180154af6;
                              }
                            }
LAB_180154afb:
                            plVar9 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803b6e08);
                            if (plVar9 == (longlong *)0x0) {
                              lVar6 = 0;
                            }
                            else {
                              lVar6 = *plVar9;
                            }
                            if (lVar6 == 0) {
                              FUN_180373340(param_1,&local_168,DAT_1803b6e08);
                              plVar13 = local_168;
                              plVar10 = local_178;
                              uVar19 = 0xbe;
                              goto LAB_1801553c2;
                            }
                          }
LAB_180154b23:
                          piVar2 = DAT_1803b6d00;
                          lVar5 = *(longlong *)(param_1 + 0x10);
                          plVar9 = *(longlong **)(lVar5 + 0x41da8);
                          if (plVar9 == (longlong *)0x0) {
                            plVar9 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,3);
                          }
                          else {
                            *(longlong *)(lVar5 + 0x41da8) = plVar9[3];
                            *(int *)(lVar5 + 0x41e40) = *(int *)(lVar5 + 0x41e40) + -1;
                            *plVar9 = 1;
                          }
                          plVar11 = plVar9 + -2;
                          plVar9[3] = 0;
                          plVar9[4] = 0;
                          plVar9[5] = 0;
                          lVar5 = _PyThreadState_GetCurrent();
                          plVar10 = local_160;
                          lVar5 = *(longlong *)(*(longlong *)(lVar5 + 0x10) + 0xd0);
                          puVar1 = *(undefined8 **)(lVar5 + 8);
                          *puVar1 = plVar11;
                          plVar9[-1] = plVar9[-1] & 3;
                          plVar9[-1] = plVar9[-1] | (ulonglong)puVar1;
                          *plVar11 = lVar5;
                          *(longlong **)(lVar5 + 8) = plVar11;
                          if (*piVar2 + 1 != 0) {
                            *piVar2 = *piVar2 + 1;
                          }
                          plVar9[3] = (longlong)piVar2;
                          lVar5 = FUN_180375520(param_1,local_160,DAT_1803b6c50);
                          if (lVar5 == 0) {
                            uVar19 = 0xbe;
                            goto LAB_18015564b;
                          }
                          plVar9[4] = lVar5;
                          piVar2 = DAT_1803b6e10;
                          if (*DAT_1803b6e10 + 1 != 0) {
                            *DAT_1803b6e10 = *DAT_1803b6e10 + 1;
                          }
                          plVar9[5] = (longlong)piVar2;
                          plVar11 = (longlong *)PyUnicode_Join(DAT_1803b6c50,plVar9);
                          if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                            (**(code **)(plVar9[1] + 0x30))(plVar9);
                          }
                          plVar9 = local_140;
                          if (plVar11 == (longlong *)0x0) {
LAB_18015562e:
                            plVar13 = *(longlong **)(param_1 + 0x60);
                            *(undefined8 *)(param_1 + 0x60) = 0;
                            plVar10 = local_178;
                            uVar19 = 0xbe;
                            goto LAB_1801553c2;
                          }
                          local_a8 = local_170;
                          *(undefined4 *)(local_140 + 5) = 0xbe;
                          local_a0 = plVar11;
                          pcVar14 = (code *)FUN_180377750(param_1,lVar6,&local_a8);
                          if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                            (**(code **)(plVar11[1] + 0x30))(plVar11);
                          }
                          if (pcVar14 == (code *)0x0) goto LAB_18015562e;
                          iVar4 = FUN_180001510(pcVar14);
                          if (iVar4 == -1) {
                            plVar13 = *(longlong **)(param_1 + 0x60);
                            *(undefined8 *)(param_1 + 0x60) = 0;
                            if ((-1 < *(int *)pcVar14) &&
                               (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1,
                               *(longlong *)pcVar14 == 0)) {
                              (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
                            }
                            plVar10 = local_178;
                            uVar19 = 0xbe;
                            goto LAB_1801553c2;
                          }
                          if (iVar4 != 1) {
                            if ((-1 < *(int *)pcVar14) &&
                               (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1,
                               *(longlong *)pcVar14 == 0)) {
                              (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
                            }
                            lVar5 = *(longlong *)(DAT_1803ca760 + 0x20);
                            if (*(char *)(lVar5 + 10) == '\0') {
                              plVar11 = (longlong *)FUN_1800015a0();
                              if (plVar11 == (longlong *)0x0) {
                                lVar6 = 0;
                              }
                              else {
                                lVar6 = *plVar11;
                              }
LAB_180154d63:
                              if (lVar6 == 0) goto LAB_180154d68;
                            }
                            else {
                              iVar4 = *(int *)(lVar5 + 0xc);
                              if (iVar4 == 0) {
                                iVar4 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
                                *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar4 + 1;
                                *(int *)(lVar5 + 0xc) = iVar4;
                              }
                              if (iVar4 != DAT_1803a98e0) {
                                DAT_1803a98e0 = iVar4;
                                DAT_1803b6a98 =
                                     FUN_1803707d0(lVar5,DAT_1803b6e18,
                                                   *(undefined8 *)(DAT_1803b6e18 + 0x18));
                              }
                              if (-1 < DAT_1803b6a98) {
                                lVar17 = (1L << (*(byte *)(lVar5 + 9) & 0x3f)) + lVar5;
                                lVar6 = *(longlong *)(lVar17 + 0x28 + DAT_1803b6a98 * 0x10);
                                if (lVar6 != 0) goto LAB_180154d90;
                                DAT_1803b6a98 =
                                     FUN_1803707d0(lVar5,DAT_1803b6e18,
                                                   *(undefined8 *)(DAT_1803b6e18 + 0x18));
                                if (-1 < DAT_1803b6a98) {
                                  lVar6 = *(longlong *)(lVar17 + 0x28 + DAT_1803b6a98 * 0x10);
                                  goto LAB_180154d63;
                                }
                              }
LAB_180154d68:
                              plVar11 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803b6e18);
                              if ((plVar11 == (longlong *)0x0) || (lVar6 = *plVar11, lVar6 == 0)) {
                                FUN_180373340(param_1,&local_168,DAT_1803b6e18);
                                plVar13 = local_168;
                                plVar10 = local_178;
                                uVar19 = 0xbf;
                                goto LAB_1801553c2;
                              }
                            }
LAB_180154d90:
                            plVar11 = local_170;
                            local_98 = local_170;
                            *(undefined4 *)(plVar9 + 5) = 0xbf;
                            local_90 = plVar10;
                            pcVar14 = (code *)FUN_180377750(param_1,lVar6,&local_98);
                            if (pcVar14 == (code *)0x0) {
                              plVar13 = *(longlong **)(param_1 + 0x60);
                              *(undefined8 *)(param_1 + 0x60) = 0;
                              plVar10 = local_178;
                              uVar19 = 0xbf;
                              goto LAB_1801553c2;
                            }
                            if (pcVar14 == _Py_TrueStruct_exref) goto LAB_180154fb3;
                            if ((pcVar14 != _Py_FalseStruct_exref) &&
                               (pcVar14 != _Py_NoneStruct_exref)) {
                              lVar5 = *(longlong *)(pcVar14 + 8);
                              if ((*(longlong *)(lVar5 + 0x60) == 0) ||
                                 (pcVar15 = *(code **)(*(longlong *)(lVar5 + 0x60) + 0x48),
                                 pcVar15 == (code *)0x0)) {
                                if ((*(undefined8 **)(lVar5 + 0x70) == (undefined8 *)0x0) ||
                                   (pcVar15 = (code *)**(undefined8 **)(lVar5 + 0x70),
                                   pcVar15 == (code *)0x0)) {
                                  if ((*(undefined8 **)(lVar5 + 0x68) == (undefined8 *)0x0) ||
                                     (pcVar15 = (code *)**(undefined8 **)(lVar5 + 0x68),
                                     pcVar15 == (code *)0x0)) goto LAB_180154fb3;
                                  lVar5 = (*pcVar15)(pcVar14);
                                }
                                else {
                                  lVar5 = (*pcVar15)(pcVar14);
                                }
                              }
                              else {
                                iVar4 = (*pcVar15)(pcVar14);
                                lVar5 = (longlong)iVar4;
                              }
                              if (0 < lVar5) goto LAB_180154fb3;
                              if (lVar5 != 0) {
                                plVar13 = *(longlong **)(param_1 + 0x60);
                                *(undefined8 *)(param_1 + 0x60) = 0;
                                if ((-1 < *(int *)pcVar14) &&
                                   (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1,
                                   *(longlong *)pcVar14 == 0)) {
                                  (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
                                }
                                plVar10 = local_178;
                                uVar19 = 0xbf;
                                goto LAB_1801553c2;
                              }
                            }
                            if ((-1 < *(int *)pcVar14) &&
                               (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1,
                               *(longlong *)pcVar14 == 0)) {
                              (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
                            }
                            lVar5 = *(longlong *)(DAT_1803ca760 + 0x20);
                            if (*(char *)(lVar5 + 10) == '\0') {
                              plVar18 = (longlong *)FUN_1800015a0();
                              if (plVar18 == (longlong *)0x0) {
                                lVar6 = 0;
                              }
                              else {
                                lVar6 = *plVar18;
                              }
LAB_180154f20:
                              if (lVar6 == 0) goto LAB_180154f25;
                            }
                            else {
                              iVar4 = *(int *)(lVar5 + 0xc);
                              if (iVar4 == 0) {
                                iVar4 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
                                *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar4 + 1;
                                *(int *)(lVar5 + 0xc) = iVar4;
                              }
                              if (iVar4 != DAT_1803a990c) {
                                DAT_1803a990c = iVar4;
                                DAT_1803b6ae8 =
                                     FUN_1803707d0(lVar5,DAT_1803b6e20,
                                                   *(undefined8 *)(DAT_1803b6e20 + 0x18));
                              }
                              if (-1 < DAT_1803b6ae8) {
                                lVar17 = (1L << (*(byte *)(lVar5 + 9) & 0x3f)) + lVar5;
                                lVar6 = *(longlong *)(lVar17 + 0x28 + DAT_1803b6ae8 * 0x10);
                                if (lVar6 != 0) goto LAB_180154f4d;
                                DAT_1803b6ae8 =
                                     FUN_1803707d0(lVar5,DAT_1803b6e20,
                                                   *(undefined8 *)(DAT_1803b6e20 + 0x18));
                                if (-1 < DAT_1803b6ae8) {
                                  lVar6 = *(longlong *)(lVar17 + 0x28 + DAT_1803b6ae8 * 0x10);
                                  goto LAB_180154f20;
                                }
                              }
LAB_180154f25:
                              plVar18 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803b6e20);
                              if ((plVar18 == (longlong *)0x0) || (lVar6 = *plVar18, lVar6 == 0)) {
                                FUN_180373340(param_1,&local_168,DAT_1803b6e20);
                                plVar13 = local_168;
                                plVar10 = local_178;
                                uVar19 = 0xc0;
                                goto LAB_1801553c2;
                              }
                            }
LAB_180154f4d:
                            *(undefined4 *)(plVar9 + 5) = 0xc0;
                            local_68 = plVar11;
                            local_60 = plVar10;
                            pcVar14 = (code *)FUN_180377750(param_1,lVar6,&local_68);
                            if (pcVar14 == (code *)0x0) {
                              plVar13 = *(longlong **)(param_1 + 0x60);
                              *(undefined8 *)(param_1 + 0x60) = 0;
                              plVar10 = local_178;
                              uVar19 = 0xc0;
                              goto LAB_1801553c2;
                            }
                            iVar4 = FUN_180001510(pcVar14);
                            if (iVar4 == -1) {
                              plVar13 = *(longlong **)(param_1 + 0x60);
                              *(undefined8 *)(param_1 + 0x60) = 0;
                              if ((-1 < *(int *)pcVar14) &&
                                 (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1,
                                 *(longlong *)pcVar14 == 0)) {
                                (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
                              }
                              plVar10 = local_178;
                              uVar19 = 0xc0;
                              goto LAB_1801553c2;
                            }
                            if (iVar4 != 1) {
                              if ((-1 < *(int *)pcVar14) &&
                                 (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1,
                                 *(longlong *)pcVar14 == 0)) {
                                (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
                              }
                              pcVar14 = DAT_1803b6c50;
                              if (*(int *)DAT_1803b6c50 + 1 != 0) {
                                *(int *)DAT_1803b6c50 = *(int *)DAT_1803b6c50 + 1;
                              }
                            }
                          }
LAB_180154fb3:
                          *(undefined4 *)(plVar9 + 5) = 0xbd;
                          plVar9 = (longlong *)FUN_180376770(param_1,lVar16,pcVar14);
                          if ((-1 < *(int *)pcVar14) &&
                             (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1,
                             *(longlong *)pcVar14 == 0)) {
                            (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
                          }
                          if (plVar9 == (longlong *)0x0) {
                            plVar13 = *(longlong **)(param_1 + 0x60);
                            *(undefined8 *)(param_1 + 0x60) = 0;
                            plVar10 = local_178;
                            uVar19 = 0xbd;
                            goto LAB_1801553c2;
                          }
                          plVar11 = plVar9;
                          if (((local_110 != (longlong *)0x0) && (-1 < (int)*local_110)) &&
                             (*local_110 = *local_110 + -1, *local_110 == 0)) {
                            plVar11 = local_110 + 1;
                            local_110 = plVar9;
                            (**(code **)(*plVar11 + 0x30))();
                            plVar11 = local_110;
                          }
                          local_110 = plVar11;
                          lVar5 = DAT_1803ca760;
                          if (local_138 == (longlong *)0x0) {
                            plVar9 = (longlong *)
                                     PyUnicode_FromFormat
                                               ("cannot access local variable \'%U\' where it is not associated with a value"
                                                ,DAT_1803b6e28);
                            plVar13 = (longlong *)
                                      FUN_180376770(param_1,*(undefined8 *)
                                                             PyExc_UnboundLocalError_exref,plVar9);
                            if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                              (**(code **)(plVar9[1] + 0x30))(plVar9);
                            }
                            FUN_18000d1d0(param_1,plVar13);
                            plVar10 = local_178;
                            uVar19 = 0xc3;
                            goto LAB_1801553c2;
                          }
                          lVar6 = *(longlong *)(DAT_1803ca760 + 0x20);
                          if (*(char *)(lVar6 + 10) == '\0') {
                            plVar11 = (longlong *)FUN_1800015a0(DAT_1803ca760,DAT_1803b6e30);
                            lVar5 = DAT_1803ca760;
                            if (plVar11 == (longlong *)0x0) {
                              lVar16 = 0;
                            }
                            else {
                              lVar16 = *plVar11;
                            }
LAB_1801550ed:
                            if (lVar16 == 0) goto LAB_1801550f2;
                          }
                          else {
                            iVar4 = *(int *)(lVar6 + 0xc);
                            if (iVar4 == 0) {
                              iVar4 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
                              *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar4 + 1;
                              *(int *)(lVar6 + 0xc) = iVar4;
                            }
                            if (iVar4 != DAT_1803a98f0) {
                              DAT_1803a98f0 = iVar4;
                              DAT_1803b6a90 =
                                   FUN_1803707d0(lVar6,DAT_1803b6e30,
                                                 *(undefined8 *)(DAT_1803b6e30 + 0x18));
                            }
                            if (-1 < DAT_1803b6a90) {
                              lVar17 = (1L << (*(byte *)(lVar6 + 9) & 0x3f)) + lVar6;
                              lVar16 = *(longlong *)(lVar17 + 0x28 + DAT_1803b6a90 * 0x10);
                              if (lVar16 != 0) goto LAB_180155121;
                              DAT_1803b6a90 =
                                   FUN_1803707d0(lVar6,DAT_1803b6e30,
                                                 *(undefined8 *)(DAT_1803b6e30 + 0x18));
                              if (-1 < DAT_1803b6a90) {
                                lVar16 = *(longlong *)(lVar17 + 0x28 + DAT_1803b6a90 * 0x10);
                                goto LAB_1801550ed;
                              }
                            }
LAB_1801550f2:
                            plVar11 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803b6e30);
                            if ((plVar11 == (longlong *)0x0) ||
                               (lVar16 = *plVar11, lVar5 = DAT_1803ca760, lVar16 == 0)) {
                              FUN_180373340(param_1,&local_168,DAT_1803b6e30);
                              plVar13 = local_168;
                              plVar10 = local_178;
                              uVar19 = 0xc3;
                              goto LAB_1801553c2;
                            }
                          }
LAB_180155121:
                          lVar6 = *(longlong *)(lVar5 + 0x20);
                          if (*(char *)(lVar6 + 10) == '\0') {
                            plVar11 = (longlong *)FUN_1800015a0(lVar5);
                            if (plVar11 == (longlong *)0x0) {
                              lVar5 = 0;
                            }
                            else {
                              lVar5 = *plVar11;
                            }
LAB_1801551e5:
                            if (lVar5 == 0) goto LAB_1801551ea;
                          }
                          else {
                            iVar4 = *(int *)(lVar6 + 0xc);
                            if (iVar4 == 0) {
                              iVar4 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
                              *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar4 + 1;
                              *(int *)(lVar6 + 0xc) = iVar4;
                            }
                            if (iVar4 != DAT_1803a991c) {
                              DAT_1803a991c = iVar4;
                              DAT_1803b6af0 =
                                   FUN_1803707d0(lVar6,DAT_1803b6e38,
                                                 *(undefined8 *)(DAT_1803b6e38 + 0x18));
                            }
                            if (-1 < DAT_1803b6af0) {
                              lVar17 = (1L << (*(byte *)(lVar6 + 9) & 0x3f)) + lVar6;
                              lVar5 = *(longlong *)(lVar17 + 0x28 + DAT_1803b6af0 * 0x10);
                              if (lVar5 != 0) goto LAB_180155212;
                              DAT_1803b6af0 =
                                   FUN_1803707d0(lVar6,DAT_1803b6e38,
                                                 *(undefined8 *)(DAT_1803b6e38 + 0x18));
                              if (-1 < DAT_1803b6af0) {
                                lVar5 = *(longlong *)(lVar17 + 0x28 + DAT_1803b6af0 * 0x10);
                                goto LAB_1801551e5;
                              }
                            }
LAB_1801551ea:
                            plVar11 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803b6e38);
                            if ((plVar11 == (longlong *)0x0) || (lVar5 = *plVar11, lVar5 == 0)) {
                              FUN_180373340(param_1,&local_168,DAT_1803b6e38);
                              plVar13 = local_168;
                              plVar10 = local_178;
                              uVar19 = 199;
                              goto LAB_1801553c2;
                            }
                          }
LAB_180155212:
                          plVar11 = local_140;
                          *(undefined4 *)(local_140 + 5) = 199;
                          plVar18 = (longlong *)FUN_180376770(param_1,lVar5,plVar9);
                          if (plVar18 == (longlong *)0x0) {
                            plVar13 = *(longlong **)(param_1 + 0x60);
                            *(undefined8 *)(param_1 + 0x60) = 0;
                            plVar10 = local_178;
                            uVar19 = 199;
                            goto LAB_1801553c2;
                          }
                          *(undefined4 *)(plVar11 + 5) = 0xc3;
                          local_80 = local_150;
                          local_88 = plVar10;
                          local_78 = plVar9;
                          local_70 = plVar18;
                          lVar5 = FUN_18037ccb0(param_1,lVar16,&local_88,DAT_1803b6e40);
                          if ((-1 < (int)*plVar18) && (*plVar18 = *plVar18 + -1, *plVar18 == 0)) {
                            (**(code **)(plVar18[1] + 0x30))(plVar18);
                          }
                          plVar9 = local_138;
                          if (lVar5 == 0) {
                            plVar13 = *(longlong **)(param_1 + 0x60);
                            *(undefined8 *)(param_1 + 0x60) = 0;
                            plVar10 = local_178;
                            uVar19 = 0xc3;
                            goto LAB_1801553c2;
                          }
                          lVar6 = local_138[2];
                          cVar3 = FUN_1803722b0(local_138,lVar6 + 1);
                          if (cVar3 != '\0') {
                            *(longlong *)(plVar9[3] + lVar6 * 8) = lVar5;
                          }
                          lVar5 = *(longlong *)(param_1 + 0x10);
                          if ((((*(int *)(*(longlong *)(lVar5 + 0x60) + 0x4d4) != 0) ||
                               (*(int *)(lVar5 + 0x1a0) != 0)) &&
                              (iVar4 = Py_MakePendingCalls(), iVar4 < 0)) &&
                             (*(longlong *)(param_1 + 0x60) != 0)) {
LAB_18015552f:
                            plVar13 = *(longlong **)(param_1 + 0x60);
                            *(undefined8 *)(param_1 + 0x60) = 0;
                            plVar10 = local_178;
                            uVar19 = 0xb6;
                            goto LAB_1801553c2;
                          }
                          if (*(int *)(lVar5 + 0x174) != 0) {
                            PyEval_SaveThread();
                            PyEval_AcquireThread(param_1);
                          }
                          plVar9 = local_158;
                          if (*(longlong *)(param_1 + 0x80) != 0) {
                            plVar9 = *(longlong **)(param_1 + 0x60);
                            *(undefined8 *)(param_1 + 0x80) = 0;
                            uVar12 = FUN_180373490(param_1);
                            *(undefined8 *)(param_1 + 0x60) = uVar12;
                            if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) &&
                               (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                              (**(code **)(plVar9[1] + 0x30))(plVar9);
                            }
                            goto LAB_18015552f;
                          }
                          local_160 = (longlong *)(**(code **)(local_158[1] + 0xe0))(local_158);
                          plVar11 = local_140;
                          plVar10 = local_178;
                        } while (local_160 != (longlong *)0x0);
                      }
                      plVar13 = *(longlong **)(param_1 + 0x60);
                      if (plVar13 != (longlong *)0x0) {
                        plVar18 = (longlong *)plVar13[1];
                        if ((*(uint *)((longlong *)plVar13[1] + 0x15) & 0x40000000) == 0) {
                          plVar18 = plVar13;
                        }
                        if ((plVar18 != *(longlong **)PyExc_StopIteration_exref) &&
                           (((-1 < *(int *)(plVar18[1] + 0xa8) ||
                             ((*(uint *)(plVar18 + 0x15) & 0x40000000) == 0)) ||
                            (cVar3 = FUN_180382640(), cVar3 != '\x01')))) {
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          uVar19 = 0xb6;
LAB_1801553c2:
                          if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
                             (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                            (**(code **)(plVar10[1] + 0x30))(plVar10);
                          }
                          plVar9 = local_170;
                          if ((-1 < (int)*local_158) &&
                             (*local_158 = *local_158 + -1, *local_158 == 0)) {
                            (**(code **)(local_158[1] + 0x30))();
                            plVar9 = local_170;
                          }
                          goto LAB_180155b57;
                        }
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                          (**(code **)(plVar13[1] + 0x30))(plVar13);
                        }
                      }
                      if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
                         (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                        (**(code **)(plVar10[1] + 0x30))(plVar10);
                      }
                      if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                        (**(code **)(plVar9[1] + 0x30))(plVar9);
                      }
                      lVar5 = *(longlong *)(DAT_1803ca760 + 0x20);
                      if (*(char *)(lVar5 + 10) == '\0') {
                        plVar9 = (longlong *)FUN_1800015a0(DAT_1803ca760,DAT_1803b6e48);
                        if (plVar9 == (longlong *)0x0) {
                          lVar6 = 0;
                        }
                        else {
                          lVar6 = *plVar9;
                        }
LAB_180155930:
                        if (lVar6 == 0) goto LAB_180155935;
                      }
                      else {
                        iVar4 = *(int *)(lVar5 + 0xc);
                        if (iVar4 == 0) {
                          iVar4 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
                          *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar4 + 1;
                          *(int *)(lVar5 + 0xc) = iVar4;
                        }
                        if (iVar4 != DAT_1803a98f4) {
                          DAT_1803a98f4 = iVar4;
                          DAT_1803b6a60 =
                               FUN_1803707d0(lVar5,DAT_1803b6e48,
                                             *(undefined8 *)(DAT_1803b6e48 + 0x18));
                        }
                        if (-1 < DAT_1803b6a60) {
                          lVar16 = (1L << (*(byte *)(lVar5 + 9) & 0x3f)) + lVar5;
                          lVar6 = *(longlong *)(lVar16 + 0x28 + DAT_1803b6a60 * 0x10);
                          if (lVar6 != 0) goto LAB_180155979;
                          DAT_1803b6a60 =
                               FUN_1803707d0(lVar5,DAT_1803b6e48,
                                             *(undefined8 *)(DAT_1803b6e48 + 0x18));
                          if (-1 < DAT_1803b6a60) {
                            lVar6 = *(longlong *)(lVar16 + 0x28 + DAT_1803b6a60 * 0x10);
                            goto LAB_180155930;
                          }
                        }
LAB_180155935:
                        plVar9 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803b6e48);
                        if ((plVar9 == (longlong *)0x0) || (lVar6 = *plVar9, lVar6 == 0)) {
                          FUN_180373340(param_1,&local_168,DAT_1803b6e48);
                          uVar19 = 0xca;
                          plVar13 = local_168;
                          plVar9 = local_170;
                          goto LAB_180155b57;
                        }
                      }
LAB_180155979:
                      plVar9 = (longlong *)FUN_18037f0c0();
                      uVar19 = 0xcc;
                      if (plVar9 == (longlong *)0x0) {
                        plVar13 = *(longlong **)(param_1 + 0x60);
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        plVar9 = local_170;
                      }
                      else {
                        *(undefined4 *)(plVar11 + 5) = 0xcc;
                        plVar18 = (longlong *)FUN_180376210(param_1,plVar9);
                        if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                          (**(code **)(plVar9[1] + 0x30))(plVar9);
                        }
                        plVar9 = local_170;
                        if (plVar18 == (longlong *)0x0) {
                          plVar13 = *(longlong **)(param_1 + 0x60);
                          *(undefined8 *)(param_1 + 0x60) = 0;
                        }
                        else if (local_138 == (longlong *)0x0) {
                          if ((-1 < (int)*plVar18) && (*plVar18 = *plVar18 + -1, *plVar18 == 0)) {
                            (**(code **)(plVar18[1] + 0x30))(plVar18);
                          }
                          plVar9 = (longlong *)
                                   PyUnicode_FromFormat
                                             ("cannot access local variable \'%U\' where it is not associated with a value"
                                              ,DAT_1803b6e28);
                          plVar13 = (longlong *)
                                    FUN_180376770(param_1,*(undefined8 *)
                                                           PyExc_UnboundLocalError_exref,plVar9);
                          if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                            (**(code **)(plVar9[1] + 0x30))(plVar9);
                          }
                          FUN_18000d1d0(param_1,plVar13);
                          uVar19 = 0xcd;
                          plVar9 = local_170;
                        }
                        else if (local_170 == (longlong *)0x0) {
                          if ((-1 < (int)*plVar18) && (*plVar18 = *plVar18 + -1, *plVar18 == 0)) {
                            (**(code **)(plVar18[1] + 0x30))(plVar18);
                          }
                          plVar9 = (longlong *)
                                   PyUnicode_FromFormat
                                             ("cannot access local variable \'%U\' where it is not associated with a value"
                                              ,DAT_1803b6df0);
                          plVar13 = (longlong *)
                                    FUN_180376770(param_1,*(undefined8 *)
                                                           PyExc_UnboundLocalError_exref,plVar9);
                          if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                            (**(code **)(plVar9[1] + 0x30))(plVar9);
                          }
                          FUN_18000d1d0(param_1,plVar13);
                          uVar19 = 0xce;
                          plVar9 = local_170;
                        }
                        else {
                          local_68 = local_130;
                          uVar19 = 0xca;
                          *(undefined4 *)(plVar11 + 5) = 0xca;
                          local_50 = local_170;
                          local_58 = local_138;
                          local_60 = plVar18;
                          lVar5 = FUN_18037ccb0(param_1,lVar6,&local_68,DAT_1803b6e50);
                          if ((-1 < (int)*plVar18) && (*plVar18 = *plVar18 + -1, *plVar18 == 0)) {
                            (**(code **)(plVar18[1] + 0x30))(plVar18);
                          }
                          if (lVar5 != 0) {
                            FUN_180001d80(param_1);
                            if ((-1 < (int)*local_138) &&
                               (*local_138 = *local_138 + -1, *local_138 == 0)) {
                              (**(code **)(local_138[1] + 0x30))();
                            }
                            if ((-1 < (int)*local_130) &&
                               (*local_130 = *local_130 + -1, *local_130 == 0)) {
                              (**(code **)(local_130[1] + 0x30))();
                            }
                            if ((-1 < (int)*local_128) &&
                               (*local_128 = *local_128 + -1, *local_128 == 0)) {
                              (**(code **)(local_128[1] + 0x30))();
                            }
                            if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
                               (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                              (**(code **)(plVar10[1] + 0x30))(plVar10);
                            }
                            if (((local_120 != (longlong *)0x0) && (-1 < (int)*local_120)) &&
                               (*local_120 = *local_120 + -1, *local_120 == 0)) {
                              (**(code **)(local_120[1] + 0x30))();
                            }
                            if (((local_110 != (longlong *)0x0) && (-1 < (int)*local_110)) &&
                               (*local_110 = *local_110 + -1, *local_110 == 0)) {
                              (**(code **)(local_110[1] + 0x30))();
                            }
                            if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                              (**(code **)(plVar9[1] + 0x30))(plVar9);
                            }
                            if ((int)*local_108 < 0) {
                              return lVar5;
                            }
                            *local_108 = *local_108 + -1;
                            if (*local_108 != 0) {
                              return lVar5;
                            }
                            (**(code **)(local_108[1] + 0x30))();
                            return lVar5;
                          }
                          plVar13 = *(longlong **)(param_1 + 0x60);
                          *(undefined8 *)(param_1 + 0x60) = 0;
                        }
                      }
                      goto LAB_180155b57;
                    }
                  }
                }
                uVar19 = 0xb6;
                goto LAB_180155b4a;
              }
              FUN_180373340(param_1,&local_168,DAT_1803b6da0);
              uVar19 = 0xb1;
              plVar13 = local_168;
              plVar9 = local_170;
              plVar10 = plVar18;
            }
          }
        }
        else {
          if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
            (**(code **)(plVar7[1] + 0x30))(plVar7);
          }
          lVar6 = FUN_18014bf10(param_1);
          if (lVar6 == 0) {
            FUN_180373340(param_1,&local_168,DAT_1803b6db8);
            uVar19 = 0xad;
            plVar13 = local_168;
            plVar10 = plVar18;
          }
          else {
            uVar19 = 0xad;
            *(undefined4 *)(plVar11 + 5) = 0xad;
            plVar7 = (longlong *)FUN_180376770(param_1,lVar6,plVar9);
            if (plVar7 != (longlong *)0x0) {
              iVar4 = FUN_180001510(plVar7);
              if (iVar4 == -1) goto LAB_180153ff1;
              if (iVar4 != 1) {
                if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                  (**(code **)(plVar7[1] + 0x30))(plVar7);
                }
                plVar7 = DAT_1803b6dc0;
                iVar4 = (int)*DAT_1803b6dc0 + 1;
                if (iVar4 != 0) {
                  *(int *)DAT_1803b6dc0 = iVar4;
                }
              }
              goto LAB_1801540da;
            }
            plVar13 = *(longlong **)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
            plVar10 = plVar18;
          }
        }
      }
    }
  }
LAB_180155b57:
  piVar2 = (int *)plVar13[5];
  if (piVar2 == (int *)0x0) {
    pcVar14 = (code *)FUN_18038e2a0(local_140,uVar19);
    plVar11 = (longlong *)plVar13[5];
    pcVar15 = pcVar14;
    if (pcVar14 == _Py_NoneStruct_exref) {
      pcVar15 = (code *)0x0;
    }
    else {
LAB_180155bb6:
      pcVar14 = pcVar15;
      if ((pcVar15 != (code *)0x0) && (*(int *)pcVar15 + 1 != 0)) {
        *(int *)pcVar15 = *(int *)pcVar15 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar2 + 6) == local_140) goto LAB_180155bff;
    pcVar14 = (code *)FUN_18038e2a0(local_140,uVar19);
    *(int **)(pcVar14 + 0x10) = piVar2;
    if (*piVar2 + 1 != 0) {
      *piVar2 = *piVar2 + 1;
    }
    plVar11 = (longlong *)plVar13[5];
    pcVar15 = pcVar14;
    if (pcVar14 != _Py_NoneStruct_exref) goto LAB_180155bb6;
    pcVar15 = (code *)0x0;
  }
  plVar13[5] = (longlong)pcVar15;
  if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
    (**(code **)(plVar11[1] + 0x30))();
  }
  if (((pcVar14 != (code *)0x0) && (-1 < *(int *)pcVar14)) &&
     (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1, *(longlong *)pcVar14 == 0)) {
    (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
  }
LAB_180155bff:
  plVar7 = local_110;
  plVar18 = local_138;
  local_180 = local_110;
  local_188 = local_120;
  local_198 = local_128;
  local_1a0 = local_130;
  local_1a8 = local_138;
  local_190 = plVar10;
  FUN_18036d0a0(local_140,"oooooooo",plVar9,local_108);
  plVar11 = DAT_1803b6b78;
  if (local_140 == DAT_1803b6b78) {
    if ((-1 < (int)*DAT_1803b6b78) && (*DAT_1803b6b78 = *DAT_1803b6b78 + -1, *plVar11 == 0)) {
      (**(code **)(plVar11[1] + 0x30))();
    }
    DAT_1803b6b78 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((plVar18 != (longlong *)0x0) && (-1 < (int)*plVar18)) &&
     (*plVar18 = *plVar18 + -1, *plVar18 == 0)) {
    (**(code **)(plVar18[1] + 0x30))(plVar18);
  }
  if (((local_130 != (longlong *)0x0) && (-1 < (int)*local_130)) &&
     (*local_130 = *local_130 + -1, *local_130 == 0)) {
    (**(code **)(local_130[1] + 0x30))();
  }
  if (((local_128 != (longlong *)0x0) && (-1 < (int)*local_128)) &&
     (*local_128 = *local_128 + -1, *local_128 == 0)) {
    (**(code **)(local_128[1] + 0x30))();
  }
  if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
    (**(code **)(plVar10[1] + 0x30))(plVar10);
  }
  if (((local_120 != (longlong *)0x0) && (-1 < (int)*local_120)) &&
     (*local_120 = *local_120 + -1, *local_120 == 0)) {
    (**(code **)(local_120[1] + 0x30))();
  }
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))(plVar7);
  }
  if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
    (**(code **)(plVar9[1] + 0x30))(plVar9);
  }
  if ((-1 < (int)*local_108) && (*local_108 = *local_108 + -1, *local_108 == 0)) {
    (**(code **)(local_108[1] + 0x30))();
  }
  plVar9 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar13;
  if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)
     ) {
    (**(code **)(plVar9[1] + 0x30))();
  }
  return 0;
}
